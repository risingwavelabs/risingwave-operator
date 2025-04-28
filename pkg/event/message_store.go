// Copyright 2023 RisingWave Labs
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package event

import "sync"

// MessageStore stores the sending event messages.
type MessageStore struct {
	mu       sync.RWMutex
	messages map[string]string
}

// SetMessage sets the event message.
func (s *MessageStore) SetMessage(event string, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.messages[event] = message
}

// MessageFor gets the event message if set. It returns the event name as a default value.
func (s *MessageStore) MessageFor(event string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if msg, ok := s.messages[event]; ok {
		return msg
	}

	return event
}

// IsMessageSet checks if message set for the given event.
func (s *MessageStore) IsMessageSet(event string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.messages[event]

	return ok
}

// NewMessageStore returns a new MessageStore.
func NewMessageStore() *MessageStore {
	return &MessageStore{
		messages: make(map[string]string),
	}
}
