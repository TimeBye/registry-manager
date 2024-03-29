// Copyright © 2019 TimeBye zhongziling@vip.qq.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import "net/url"

type Registry struct {
	Uri          *url.URL
	Url          string   `mapstructure:"registry"`
	Username     string   `mapstructure:"username"`
	Password     string   `mapstructure:"password"`
	Insecure     bool     `mapstructure:"insecure"`
	Repositories []string `mapstructure:"repositories"`
}

type Config struct {
	SyncPolicy   SyncPolicy           `mapstructure:"sync-policy"`
	DeletePolicy DeletePolicy         `mapstructure:"delete-policy"`
	Registries   map[string]*Registry `mapstructure:"registries"`
}
