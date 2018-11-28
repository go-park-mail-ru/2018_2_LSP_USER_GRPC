package main

import (
	"database/sql"
	"fmt"
	"net"
	"os"

	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user"
	"github.com/go-park-mail-ru/2018_2_LSP_USER_GRPC/user_proto"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Println("Can't create logger", err)
		return
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	connStr := "host=" + os.Getenv("DB_HOST") + " user=" + os.Getenv("DB_USER") + " password=" + os.Getenv("DB_PASS") + " dbname=" + os.Getenv("DB_DB") + " sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Println("Error:", err.Error())
		return
	}
	defer db.Close()

	server := grpc.NewServer()
	user_proto.RegisterUserCheckerServer(server, user.NewUserManager(db, sugar))

	sugar.Infow("Starting server",
		"port", 8080,
	)
	lis, err := net.Listen("tcp", ":8080")
	if err != nil {
		sugar.Errorw("Can't create server",
			"port", 8080,
		)
		return
	}
	server.Serve(lis)
}
