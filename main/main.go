package main

import (
	"github.com/echsylon/go-args"
	"github.com/echsylon/go-log"
)

func main() {
	log.SetLogLevel(log.LOG_LEVEL_INFORMATIONAL)
	log.SetLogColumnSeparator(" ")
	log.SetLogColumns(
		log.LOG_COLUMN_DATETIME,
		log.LOG_COLUMN_PID,
		log.LOG_COLUMN_SOURCE,
		log.LOG_COLUMN_LEVEL,
	)

	args.SetApplicationDescription("This application enables distributed store features in a network.")
	args.DefineOptionStrict("m", "message-port", "The network messages port. Default: 8881", `^[0-9]{4,5}$`)
	args.DefineOptionStrict("r", "request-port", "The REST API port. Default: 8880", `^[0-9]{4,5}$`)
	args.DefineOptionHelp("h", "help", "Prints this help text.")

	args.Parse()

	httpPort := args.GetOptionIntValue("r", 8880)
	udpPort := args.GetOptionIntValue("m", 8881)

	controller := NewController()
	controller.SetupInfrastructure(int(httpPort), int(udpPort))
	controller.StartApiServer()
}
