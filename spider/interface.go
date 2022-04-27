package spider

type Spider interface {
	AddQueue(url string) error
	Run()
}
