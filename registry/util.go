package registry

// CopyService make a copy of service
func CopyService(service *Service) *Service {
	// copy service
	s := new(Service)
	*s = *service

	// copy nodes
	methods := make([]*Method, len(service.Methods))
	for i, method := range service.Methods {
		m := new(Method)
		*m = *method

		// copy bindings
		bindings := make([]*Binding, len(method.Bindings))
		for j, route := range method.Bindings {
			r := new(Binding)
			*r = *route

			p := new(PathTmpl)
			*p = *route.PathTmpl
			r.PathTmpl = p

			bindings[j] = r
		}
		m.Bindings = bindings

		methods[i] = m
	}
	s.Methods = methods

	return s
}

// Copy makes a copy of services
func Copy(current []*Service) []*Service {
	services := make([]*Service, len(current))
	for i, service := range current {
		services[i] = CopyService(service)
	}
	return services
}
