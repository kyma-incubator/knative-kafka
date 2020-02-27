package health

// Constants
const (
	// Default Health Configuration
	HealthPort    = 8082        // The Port That Serves The Liveness And Readiness Endpoints
	LivenessPath  = "/healthz"  // The Endpoint Of The Liveness Check
	ReadinessPath = "/healthy"  // The Endpoint Of The Readiness Check
)
