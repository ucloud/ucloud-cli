package base

import (
	"bytes"
	"io/ioutil"
	"testing"
)

var cliConfigJSON = `[{"project_id":"org-bdks4e","region":"cn-bj2","zone":"cn-bj2-04","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"uweb","active":true},{"project_id":"org-oxjwoi","region":"hk","zone":"hk-02","base_url":"https://api.ucloud.cn/","timeout_sec":15,"profile":"test","active":false}]`

var credentialJSON = `[{"public_key":"4E9UU*****3ZAPWQ==","private_key":"6945*****a0d45","profile":"uweb"},{"public_key":"YSQG*****zgnCRQ=","private_key":"jtma*****Avms","profile":"test"}]`

func TestAggConfigManager(t *testing.T) {
	cfgFilePath := GetConfigDir() + "/config.tmp.json"
	credFilePath := GetConfigDir() + "/credentail.tmp.json"

	err := ioutil.WriteFile(cfgFilePath, []byte(cliConfigJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}
	err = ioutil.WriteFile(credFilePath, []byte(credentialJSON), 0644)
	if err != nil {
		t.Fatal(err)
	}

	acManager := &AggConfigManager{
		configs:  make(map[string]*AggConfig),
		cfgFile:  bytes.NewBufferString(cliConfigJSON),
		credFile: bytes.NewBufferString(credentialJSON),
	}

	err = acManager.Load()
	if err != nil {
		t.Fatal(err)
	}

	if len(acManager.configs) != 2 {
		t.Errorf("expect length of configs is 2, accpet %d", len(acManager.configs))
	}
}
