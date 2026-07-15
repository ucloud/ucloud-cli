package pgsql

// UPgSQL-domain state constants.
//
// These follow the LIVE API State enum (as observed from ListUPgSQLInstance /
// GetUPgSQLInstance responses): Initing / InitFailed / Starting / Running /
// Stopping / Stopped / Deleted / Upgrading / Promoting / Recovering /
// RecoverFailed / StartFailed / ShutdownFailed / Deleting / DeleteFailed.
//
// NOTE: the published GetUPgSQLInstance doc lists "Shutoff"/"Shutdown"/"Fail"
// but the live API does NOT return those — a stopped instance reports State=
// "Stopped" (not "Shutoff"). Using the doc values caused the stop poller to
// never match its target and spin until the 10m timeout. Trust the live enum.
const (
	PGSQL_RUNNING         = "Running"
	PGSQL_STOPPING        = "Stopping"
	PGSQL_STOPPED         = "Stopped"
	PGSQL_INITING         = "Initing"
	PGSQL_INIT_FAILED     = "InitFailed"
	PGSQL_STARTING        = "Starting"
	PGSQL_START_FAILED    = "StartFailed"
	PGSQL_SHUTDOWN_FAILED = "ShutdownFailed"
	PGSQL_DELETING        = "Deleting"
	PGSQL_DELETED         = "Deleted"
	PGSQL_DELETE_FAILED   = "DeleteFailed"
	PGSQL_UPGRADING       = "Upgrading"
	PGSQL_PROMOTING       = "Promoting"
	PGSQL_RECOVERING      = "Recovering"
	PGSQL_RECOVER_FAILED  = "RecoverFailed"
)

// Backup state constants, from ListUPgSQLBackup.UPgSQLBackup.State enum values
// (Backuping / Success / Failed / Expired). Display-only; backup ops are not polled.
const (
	PGSQL_BACKUP_SUCCESS = "Success"
	PGSQL_BACKUP_FAILED  = "Failed"
)

// USupabase state constants (live-observed from ListUSupabaseInstance /
// DescribeUSupabase responses). Used as poll targets for start/stop/restart.
// As with the pgsql enum, the published doc is untrusted; these are confirmed
// live (Running confirmed; Stopped pending a live stop verification).
const (
	SUPABASE_RUNNING = "Running"
	SUPABASE_STOPPED = "Stopped"
	SUPABASE_FAIL    = "Fail"
)
