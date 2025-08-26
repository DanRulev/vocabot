package client

type Clients struct {
	*MyMemoryAPI
	*PythonAnyWhereAPI
	*VercelAPI
}

func InitClients() Clients {
	return Clients{
		MyMemoryAPI:       NewMyMemoryAPI(),
		PythonAnyWhereAPI: NewPythonAnyWhereAPI(),
		VercelAPI:         NewVercelAPI(),
	}
}
