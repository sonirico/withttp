package withttp

import "net/url"

func BaseURL(raw string) ReqOption {
	return ReqOptionFunc(func(req Request) (err error) {
		u, err := url.Parse(raw)
		if err != nil {
			return err
		}

		req.SetURL(u)

		return
	})
}
