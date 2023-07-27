package orchestrator

func GetDriver(driverType string) (Driver, error) {
	switch driverType {
	case DriverTypePrinter:
		return NewK8sPrinterDriver(), nil
	default:
		return NewK8sDriver("")
	}
}
