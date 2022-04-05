// Copyright 2021 The Karavel Project
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logger

import "context"

type contextKey string

var logKey contextKey = "logger.log"

func WithLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, logKey, log)
}

func FromContext(ctx context.Context) Logger {
	logger, ok := ctx.Value(logKey).(Logger)
	if !ok || logger == nil {
		// Create logger on-the-fly, but with a warning message
		logger = New(LvlInfo)
		logger.Warn("tried to retrieve logger from context but none was found")
	}
	return logger
}
