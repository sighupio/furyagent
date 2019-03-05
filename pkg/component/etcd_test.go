package component

func TestTimeConsuming(t *testing.T) {
	agentConfig, _ := InitAgent("furyagent.yml")
	store, _ := storage.Init(&agentConfig.Storage)
	cfg, _ := getEtcdCfg(c, store)
	cli, _ := clientv3.New(*cfg)
	defer cli.Close()

}
