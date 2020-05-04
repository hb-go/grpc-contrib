package registry

type MockRegistry struct {
}

func (*MockRegistry) Init(...Option) error {
	return nil
}

func (*MockRegistry) Options() Options {
	return Options{}
}

func (*MockRegistry) NewTarget(*Service, ...Option) string {
	return ""
}
func (*MockRegistry) Register(*Service, ...RegisterOption) error {
	return nil
}
func (*MockRegistry) Deregister(*Service) error {
	return nil
}

func (*MockRegistry) GetService(string) ([]*Service, error) {
	return []*Service{}, nil
}

func (*MockRegistry) ListServices() ([]*Service, error) {
	return []*Service{}, nil
}

func (*MockRegistry) Watch(...WatchOption) (Watcher, error) {
	return &mockWatcher{}, nil
}

func (*MockRegistry) String() string {
	return "mock"
}

type mockWatcher struct {
}

func (*mockWatcher) Next() (*Result, error) {
	ch := make(chan bool, 1)
	<-ch
	return nil, nil
}
func (*mockWatcher) Stop() {

}
