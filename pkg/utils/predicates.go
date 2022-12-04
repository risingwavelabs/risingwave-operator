package utils

import (
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
)

var (
	// CreateEventFilter is the predicate for CreateEvent only.
	CreateEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return true
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// DeleteEventFilter is the predicate for DeleteEvent only.
	DeleteEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return true
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// UpdateEventFilter is the predicate for UpdateEvent only.
	UpdateEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return true
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return false
		},
	}

	// GenericEventFilter is the predicate for GenericEvent only.
	GenericEventFilter = predicate.Funcs{
		CreateFunc: func(event event.CreateEvent) bool {
			return false
		},
		DeleteFunc: func(event event.DeleteEvent) bool {
			return false
		},
		UpdateFunc: func(event event.UpdateEvent) bool {
			return false
		},
		GenericFunc: func(event event.GenericEvent) bool {
			return true
		},
	}
)
