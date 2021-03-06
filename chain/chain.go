// Copyright (c) 2013-2015 The btcsuite developers
// Copyright (c) 2015-2016 The Decred developers
// Use of this source code is governed by an ISC
// license that can be found in the LICENSE file.

package chain

import (
	"bytes"
	"context"
	"sync"
	"time"

	"github.com/fonero-project/fnod/chaincfg"
	"github.com/fonero-project/fnod/chaincfg/chainhash"
	fnorpcclient "github.com/fonero-project/fnod/rpcclient"
	"github.com/fonero-project/fnod/wire"
	"github.com/fonero-project/fnowallet/errors"
)

var requiredChainServerAPI = semver{major: 5, minor: 0, patch: 0}

// RPCClient represents a persistent client connection to a fonero RPC server
// for information regarding the current best block chain.
type RPCClient struct {
	*fnorpcclient.Client
	connConfig  *fnorpcclient.ConnConfig // Work around unexported field
	chainParams *chaincfg.Params

	enqueueNotification       chan interface{}
	dequeueNotification       chan interface{}
	enqueueVotingNotification chan interface{}
	dequeueVotingNotification chan interface{}

	quit    chan struct{}
	wg      sync.WaitGroup
	started bool
	quitMtx sync.Mutex
}

// NewRPCClient creates a direct client connection to the server described by the
// connect string.  If disableTLS is false, the remote RPC certificate must be
// provided in the certs slice.  The connection is not established immediately,
// but must be done using the Start method.  If the remote server does not
// operate on the same fonero network as described by the passed chain
// parameters, the connection will be disconnected.
// Deprecated: use NewRPCClientConfig
func NewRPCClient(chainParams *chaincfg.Params, connect, user, pass string, certs []byte,
	disableTLS bool) (*RPCClient, error) {
	return NewRPCClientConfig(chainParams, &fnorpcclient.ConnConfig{
		Host:                 connect,
		Endpoint:             "ws",
		User:                 user,
		Pass:                 pass,
		Certificates:         certs,
		DisableAutoReconnect: true,
		DisableConnectOnNew:  true,
		DisableTLS:           disableTLS,
	})
}

// NewRPCClientConfig creates a client connection to the server described by the
// passed chainParams and connConfig
func NewRPCClientConfig(chainParams *chaincfg.Params, connConfig *fnorpcclient.ConnConfig) (*RPCClient, error) {
	client := &RPCClient{
		connConfig:                connConfig,
		chainParams:               chainParams,
		enqueueNotification:       make(chan interface{}),
		dequeueNotification:       make(chan interface{}),
		enqueueVotingNotification: make(chan interface{}),
		dequeueVotingNotification: make(chan interface{}),
		quit:                      make(chan struct{}),
	}
	ntfnCallbacks := &fnorpcclient.NotificationHandlers{
		OnBlockConnected:        client.onBlockConnected,
		OnBlockDisconnected:     client.onBlockDisconnected,
		OnRelevantTxAccepted:    client.onRelevantTxAccepted,
		OnReorganization:        client.onReorganization,
		OnWinningTickets:        client.onWinningTickets,
		OnSpentAndMissedTickets: client.onSpentAndMissedTickets,
		OnStakeDifficulty:       client.onStakeDifficulty,
	}
	rpcClient, err := fnorpcclient.New(client.connConfig, ntfnCallbacks)
	if err != nil {
		return nil, err
	}
	client.Client = rpcClient
	return client, nil
}

// Start attempts to establish a client connection with the remote server.
// If successful, handler goroutines are started to process notifications
// sent by the server.  After a limited number of connection attempts, this
// function gives up, and therefore will not block forever waiting for the
// connection to be established to a server that may not exist.
func (c *RPCClient) Start(ctx context.Context, retry bool) (err error) {
	const op errors.Op = "rpcclient.Start"

	err = c.Client.Connect(ctx, retry)
	if err != nil {
		return errors.E(op, errors.IO, err)
	}

	defer func() {
		if err != nil {
			c.Disconnect()
		}
	}()

	// Verify that the server is running on the expected network.
	net, err := c.Client.GetCurrentNet()
	if err != nil {
		return errors.E(op, errors.E(errors.Op("fnod.jsonrpc.getcurrentnet"), err))
	}
	if net != c.chainParams.Net {
		return errors.E(op, "mismatched networks")
	}

	// Ensure the RPC server has a compatible API version.
	var serverAPI semver
	versions, err := c.Client.Version()
	if err == nil {
		versionResult := versions["fnodjsonrpcapi"]
		serverAPI = semver{
			major: versionResult.Major,
			minor: versionResult.Minor,
			patch: versionResult.Patch,
		}
	}
	if !semverCompatible(requiredChainServerAPI, serverAPI) {
		return errors.E(op, errors.Errorf("advertised API version %v incompatible "+
			"with required version %v", serverAPI, requiredChainServerAPI))
	}

	c.quitMtx.Lock()
	c.started = true
	c.quitMtx.Unlock()

	c.wg.Add(2)
	go c.handler()
	go c.handlerVoting()
	return nil
}

// Stop disconnects the client and signals the shutdown of all goroutines
// started by Start.
func (c *RPCClient) Stop() {
	c.quitMtx.Lock()
	select {
	case <-c.quit:
	default:
		close(c.quit)
		c.Client.Shutdown()

		if !c.started {
			close(c.dequeueNotification)
			close(c.dequeueVotingNotification)
		}
	}
	c.quitMtx.Unlock()
}

// WaitForShutdown blocks until both the client has finished disconnecting
// and all handlers have exited.
func (c *RPCClient) WaitForShutdown() {
	c.Client.WaitForShutdown()
	c.wg.Wait()
}

// Notification types.  These are defined here and processed from from reading
// a notificationChan to avoid handling these notifications directly in
// fnorpcclient callbacks, which isn't very Go-like and doesn't allow
// blocking client calls.
type (
	// blockConnected is a notification for a newly-attached block to the
	// best chain.
	blockConnected struct {
		blockHeader  []byte
		transactions [][]byte
	}

	// blockDisconnected is a notifcation that the block described by the header
	// was reorganized out of the best chain.
	blockDisconnected struct {
		blockHeader []byte
	}

	// relevantTxAccepted is a notification that a transaction accepted by
	// mempool passed the client's transaction filter.
	relevantTxAccepted struct {
		transaction *wire.MsgTx
	}

	// reorganization is a notification that a reorg has happen with the new
	// old and new tip included.
	reorganization struct {
		oldHash   *chainhash.Hash
		oldHeight int64
		newHash   *chainhash.Hash
		newHeight int64
	}

	// winningTickets is a notification with the winning tickets (and the
	// block they are in.
	winningTickets struct {
		blockHash   *chainhash.Hash
		blockHeight int64
		tickets     []*chainhash.Hash
	}

	// missedTickets is a notifcation for tickets that have been missed.
	missedTickets struct {
		blockHash   *chainhash.Hash
		blockHeight int64
		tickets     []*chainhash.Hash
	}

	// stakeDifficulty is a notification for the current stake difficulty.
	stakeDifficulty struct {
		blockHash   *chainhash.Hash
		blockHeight int64
		stakeDiff   int64
	}
)

// notifications returns a channel of parsed notifications sent by the remote
// fonero RPC server.  This channel must be continually read or the process
// may abort for running out memory, as unread notifications are queued for
// later reads.
func (c *RPCClient) notifications() <-chan interface{} {
	return c.dequeueNotification
}

// notificationsVoting returns a channel of parsed voting notifications sent
// by the remote RPC server.  This channel must be continually read or the
// process may abort for running out memory, as unread notifications are
// queued for later reads.
func (c *RPCClient) notificationsVoting() <-chan interface{} {
	return c.dequeueVotingNotification
}

func (c *RPCClient) onBlockConnected(header []byte, transactions [][]byte) {
	select {
	case c.enqueueNotification <- blockConnected{
		blockHeader:  header,
		transactions: transactions,
	}:
	case <-c.quit:
	}
}

func (c *RPCClient) onBlockDisconnected(header []byte) {
	select {
	case c.enqueueNotification <- blockDisconnected{
		blockHeader: header,
	}:
	case <-c.quit:
	}
}

func (c *RPCClient) onRelevantTxAccepted(transaction []byte) {
	msgTx := new(wire.MsgTx)
	err := msgTx.Deserialize(bytes.NewReader(transaction))
	if err != nil {
		log.Errorf("Failed to deserialize announced transaction: %v", err)
		return
	}
	select {
	case c.enqueueNotification <- relevantTxAccepted{
		transaction: msgTx,
	}:
	case <-c.quit:
	}
}

// onReorganization handles reorganization notifications and passes them
// downstream to the notifications queue.
func (c *RPCClient) onReorganization(oldHash *chainhash.Hash, oldHeight int32,
	newHash *chainhash.Hash, newHeight int32) {
	select {
	case c.enqueueNotification <- reorganization{
		oldHash,
		int64(oldHeight),
		newHash,
		int64(newHeight),
	}:
	case <-c.quit:
	}
}

// onWinningTickets handles winning tickets notifications data and passes it
// downstream to the notifications queue.
func (c *RPCClient) onWinningTickets(hash *chainhash.Hash, height int64, tickets []*chainhash.Hash) {
	select {
	case c.enqueueVotingNotification <- winningTickets{
		blockHash:   hash,
		blockHeight: height,
		tickets:     tickets,
	}:
	case <-c.quit:
	}
}

// onSpentAndMissedTickets handles missed tickets notifications data and passes
// it downstream to the notifications queue.
func (c *RPCClient) onSpentAndMissedTickets(blockHash *chainhash.Hash, height int64, sdiff int64, tickets map[chainhash.Hash]bool) {
	var missed []*chainhash.Hash

	// Copy the missed ticket hashes to a slice.
	for ticketHash, spent := range tickets {
		if !spent {
			ticketHash := ticketHash
			missed = append(missed, &ticketHash)
		}
	}

	if len(missed) == 0 {
		return
	}

	select {
	case c.enqueueNotification <- missedTickets{
		blockHash:   blockHash,
		blockHeight: height,
		tickets:     missed,
	}:
	case <-c.quit:
	}
}

// onStakeDifficulty handles stake difficulty notifications data and passes it
// downstream to the notification queue.
func (c *RPCClient) onStakeDifficulty(hash *chainhash.Hash,
	height int64,
	stakeDiff int64) {

	select {
	case c.enqueueNotification <- stakeDifficulty{
		hash,
		height,
		stakeDiff,
	}:
	case <-c.quit:
	}
}

// handler maintains a queue of notifications and the current state (best
// block) of the chain.
func (c *RPCClient) handler() {
	// TODO: Rather than leaving this as an unbounded queue for all types of
	// notifications, try dropping ones where a later enqueued notification
	// can fully invalidate one waiting to be processed.  For example,
	// blockconnected notifications for greater block heights can remove the
	// need to process earlier blockconnected notifications still waiting
	// here.

	var notifications []interface{}
	enqueue := c.enqueueNotification
	var dequeue chan interface{}
	var next interface{}
	pingChan := time.After(time.Minute)
	pingChanReset := make(chan (<-chan time.Time))
out:
	for {
		select {
		case n, ok := <-enqueue:
			if !ok {
				// If no notifications are queued for handling,
				// the queue is finished.
				if len(notifications) == 0 {
					break out
				}
				// nil channel so no more reads can occur.
				enqueue = nil
				continue
			}
			if len(notifications) == 0 {
				next = n
				dequeue = c.dequeueNotification
			}
			notifications = append(notifications, n)
			pingChan = time.After(time.Minute)

		case dequeue <- next:
			notifications[0] = nil
			notifications = notifications[1:]
			if len(notifications) != 0 {
				next = notifications[0]
			} else {
				// If no more notifications can be enqueued, the
				// queue is finished.
				if enqueue == nil {
					break out
				}
				dequeue = nil
			}

		case <-pingChan:
			// No notifications were received in the last 60s. Ensure the
			// connection is still active by making a new request to the server.
			//
			// This MUST wait for the response in a new goroutine so as to not
			// block channel sends enqueueing more notifications.  Doing so
			// would cause a deadlock and after the timeout expires, the client
			// would be shut down.
			//
			// TODO: A minute timeout is used to prevent the handler loop from
			// blocking here forever, but this is much larger than it needs to
			// be due to fnod processing websocket requests synchronously (see
			// https://github.com/btcsuite/btcd/issues/504).  Decrease this to
			// something saner like 3s when the above issue is fixed.
			type sessionResult struct {
				err error
			}
			sessionResponse := make(chan sessionResult, 1)
			go func() {
				_, err := c.Session()
				sessionResponse <- sessionResult{err}
			}()
			go func() {
				select {
				case resp := <-sessionResponse:
					if resp.err != nil {
						log.Errorf("Failed to receive session "+
							"result: %v", resp.err)
						c.Stop()
					}
					pingChanReset <- time.After(time.Minute)

				case <-time.After(time.Minute):
					log.Errorf("Timeout waiting for session RPC")
					c.Stop()
				}
			}()

		case ch := <-pingChanReset:
			pingChan = ch

		case <-c.quit:
			break out
		}
	}

	c.Stop()
	close(c.dequeueNotification)
	c.wg.Done()
}

// handler maintains a queue of notifications and the current state (best
// block) of the chain.
func (c *RPCClient) handlerVoting() {
	var notifications []interface{}
	enqueue := c.enqueueVotingNotification
	var dequeue chan interface{}
	var next interface{}
out:
	for {
		select {
		case n, ok := <-enqueue:
			if !ok {
				// If no notifications are queued for handling,
				// the queue is finished.
				if len(notifications) == 0 {
					break out
				}
				// nil channel so no more reads can occur.
				enqueue = nil
				continue
			}
			if len(notifications) == 0 {
				next = n
				dequeue = c.dequeueVotingNotification
			}
			notifications = append(notifications, n)

		case dequeue <- next:
			notifications[0] = nil
			notifications = notifications[1:]
			if len(notifications) != 0 {
				next = notifications[0]
			} else {
				// If no more notifications can be enqueued, the
				// queue is finished.
				if enqueue == nil {
					break out
				}
				dequeue = nil
			}

		case <-c.quit:
			break out
		}
	}
	close(c.dequeueVotingNotification)
	c.wg.Done()
}

// POSTClient creates the equivalent HTTP POST fnorpcclient.Client.
func (c *RPCClient) POSTClient() (*fnorpcclient.Client, error) {
	configCopy := *c.connConfig
	configCopy.HTTPPostMode = true
	client, err := fnorpcclient.New(&configCopy, nil)
	if err != nil {
		return nil, errors.E(errors.Op("chain.POSTClient"), err)
	}
	return client, nil
}
