package main

import (
	"github.com/huobirdcenter/huobi_golang/logging/applogger"
	hclient "github.com/huobirdcenter/huobi_golang/pkg/client"
	"github.com/nathanusask/huobi-market-data/config"
)

func main() {
	// Get the timestamp from Huobi server and print on console
	client := new(hclient.CommonClient).Init(config.Host)
	resp, err := client.GetTimestamp()

	if err != nil {
		applogger.Error("Get timestamp error: %s", err)
	} else {
		applogger.Info("Get timestamp: %d", resp)
	}

	// Get the list of accounts owned by this API user and print the detail on console
	accountClient := new(hclient.AccountClient).Init(config.AccessKey, config.SecretKey, config.Host)
	resp, err = accountClient.GetAccountInfo()
	if err != nil {
		applogger.Error("Get account error: %s", err)
	} else {
		applogger.Info("Get account, count=%d", len(resp))
		for _, result := range resp {
			applogger.Info("account: %+v", result)
		}
	}
}
