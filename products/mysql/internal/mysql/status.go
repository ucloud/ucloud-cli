package mysql

// UDB-domain state constants, product-owned copies (formerly model/status).
const (
	UDB_FAIL         = "Fail"
	UDB_RUNNING      = "Running"
	UDB_SHUTOFF      = "Shutoff"
	UDB_RECOVER_FAIL = "Recover fail"
	UDB_UPGRADE_FAIL = "UpgradeFail"
	UDB_TOBE_SWITCH  = "WaitForSwitch"
)
