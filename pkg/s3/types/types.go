/*
 * Copyright 2022 Singularity Data
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 * http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package types

import (
	"context"
	"fmt"
	"strings"
)

type Client interface {
	CreateBucket(context.Context, string) error
	DeleteBucket(context.Context, string) error
}

type CreateOption interface{}

type CreateFun func(CreateOption) (Client, error)

var s3Provider = make(map[string]CreateFun)

func RegisterClientFun(provider string, f CreateFun) {
	s3Provider[provider] = f
}

func GetClientFun(provider string) (CreateFun, error) {
	var proStr = strings.ToUpper(provider)
	c, e := s3Provider[proStr]
	if !e {
		return nil, fmt.Errorf("no s3 client function for provider %s", provider)
	}
	return c, nil
}

type Factory interface {
	CreateClient() (Client, error)
}
