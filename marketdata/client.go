package marketdata

type client struct {
}

type Client interface {
	Connect() error
}

func New() Client {
	return &client{}
}

func (c *client) Connect() error {
	return nil
}
