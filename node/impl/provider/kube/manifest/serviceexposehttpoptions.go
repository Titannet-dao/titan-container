package manifest

type ServiceExposeHTTPOptions struct {
	maxBodySize uint32
	readTimeout uint32
	SendTimeout uint32
	nextTries   uint32
	nextTimeout uint32
	nextCases   []string
}
