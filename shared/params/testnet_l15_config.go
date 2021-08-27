package params

// UseL15NetworkConfig uses the Lukso specific
// network config.
func UseL15NetworkConfig() {
	cfg := BeaconNetworkConfig().Copy()
	cfg.ContractDeploymentBlock = 0
	cfg.BootstrapNodes = []string{
		"enr:-Ku4QEL0I7H3EawRwc2ZUevmj-_T0R6JZGMhfp_2KHBlwAt5bwA19c8LSYZzy63EvpsYbifKye6qnE-_vsNimWOz8scBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCPqeliJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk",
		"enr:-Ku4QAmYtwrQBZ-WJwTPL4xMpTO6BlZcU6IuXljtd_SgC51nGRs98WvxCX0-ZJBs0G9m9tcFPsktbdSr7EliMhrZnfEBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCKNcPOJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk",
		"enr:-Ku4QEXRrSXB7od-xNeoLuq6GicTHpuuCNRPPR9tM48Iai0-FoHL4JsntmpnwUrC-di-lT6gkbxV7Jikg9s6ImsAT1oBh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCPGqi6Jc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk",
		"enr:-Ku4QBuS5wqvF6SHaPpuu4r4ZlRRVC1Ojp1zDOAVC1X0PB3gRujAhWZdk2m0kn3FwoPuHft_Sku0tWHSfBVlHoER160Bh2F0dG5ldHOIAAAAAAAAAACEZXRoMpCvIkw2g6VTF___________gmlkgnY0gmlwhCKNLsqJc2VjcDI1NmsxoQLt36VpP56n0SlTYWcSBwL7aGK_AFwNLGxOGQt91nchMYN1ZHCCEuk",
	}
	OverrideBeaconNetworkConfig(cfg)
}

// UseL15Config sets the main beacon chain
// config for Lukso.
func UseL15Config() {
	beaconConfig = L15Config()
}

// L15Config defines the config for the
// Lukso testnet.
func L15Config() *BeaconChainConfig {
	cfg := MainnetConfig().Copy()
	cfg.MinGenesisTime = 1629445728
	cfg.GenesisDelay = 0
	cfg.ConfigName = ConfigNames[L15]
	cfg.GenesisForkVersion = []byte{0x83, 0xa5, 0x53, 0x17}
	cfg.SecondsPerETH1Block = 6
	cfg.SlotsPerEpoch = 32
	cfg.SecondsPerSlot = 6
	cfg.DepositChainID = 808080
	cfg.DepositNetworkID = 808080
	cfg.DepositContractAddress = "0x000000000000000000000000000000000000cafe"
	return cfg
}
