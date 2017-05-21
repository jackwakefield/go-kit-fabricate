package addsvc

// Middleware describes a service (as opposed to endpoint) middleware.
type Middleware func(Service) Service
