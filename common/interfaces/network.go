// Copyright 2017 Factom Foundation
// Use of this source code is governed by the MIT
// license that can be found in the LICENSE file.

package interfaces

type INetwork interface {
	Recieve() IMsg
	Broadcast(IMsg)
	SendToPeer(IMsg)
	Control()
	GetMLog() IMLog
	SetMLog(IMLog)
}
