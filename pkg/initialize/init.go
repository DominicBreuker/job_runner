package initialize

func init() {
	All()
}

func All() error {
	initAWS()
	initLogs()

	return nil
}
