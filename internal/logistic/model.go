package logistic

import "time"

type ApplicationStatus string

const (
	StatusNew        ApplicationStatus = "NEW"
	StatusInProgress ApplicationStatus = "IN_PROGRESS"
	StatusShipped    ApplicationStatus = "SHIPPED"
	StatusDelivered  ApplicationStatus = "DELIVERED"
	StatusCancelled  ApplicationStatus = "CANCELLED"
)

type LogisticApplication struct {
	ID                    int64             `json:"id"`
	OriginalApplicationID int64             `json:"original_application_id"`
	Status                ApplicationStatus `json:"status"`
	CreatedAt             time.Time         `json:"created_at"`
	UpdatedAt             time.Time         `json:"updated_at"`
}

type RouteStatus string

const (
	RouteDraft      RouteStatus = "DRAFT"
	RouteScheduled  RouteStatus = "SCHEDULED"
	RouteInProgress RouteStatus = "IN_PROGRESS"
	RouteCompleted  RouteStatus = "COMPLETED"
)

type Route struct {
	ID               int64       `json:"id"`
	TruckVolume      float64     `json:"truck_volume"`
	TruckMaxWeight   float64     `json:"truck_max_weight"`
	DepartureDate    time.Time   `json:"departure_date"`
	Status           RouteStatus `json:"status"`
	CreatedByManager int64       `json:"created_by_manager_id"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
}

type RoutePoint struct {
	ID               int64     `json:"id"`
	RouteID          int64     `json:"route_id"`
	LogisticsPointID int64     `json:"logistics_point_id"`
	PointOrder       int       `json:"point_order"`
	PlannedArrival   time.Time `json:"planned_arrival"`
}

type CreateRouteRequest struct {
	TruckVolume    float64           `json:"truck_volume"`
	TruckMaxWeight float64           `json:"truck_max_weight"`
	DepartureDate  time.Time         `json:"departure_date"`
	RoutePoints    []RoutePointInput `json:"route_points"`
}

type RoutePointInput struct {
	LogisticsPointID int64     `json:"logistics_point_id"`
	PointOrder       int       `json:"point_order"`
	PlannedArrival   time.Time `json:"planned_arrival"`
}

type UpdateStatusRequest struct {
	Status ApplicationStatus `json:"status"`
}
