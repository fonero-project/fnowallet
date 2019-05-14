module github.com/fonero-project/fnowallet

require (
	github.com/boltdb/bolt v1.3.1
	github.com/decred/slog v1.0.0
	github.com/golang/protobuf v1.2.0
	github.com/gorilla/websocket v1.4.0
	github.com/jessevdk/go-flags v1.4.0
	github.com/jrick/bitset v1.0.0
	github.com/jrick/logrotate v1.0.0
	github.com/fonero-project/fnod v0.0.0-20190514051128-48e51dacc4c3
	golang.org/x/crypto v0.0.0-20190222235706-ffb98f73852f
	golang.org/x/net v0.0.0-20190213061140-3a22650c66bd
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4
	google.golang.org/genproto v0.0.0-20190219182410-082222b4a5c5 // indirect
	google.golang.org/grpc v1.18.0
)

replace github.com/fonero-project/fnod => ../fnod
