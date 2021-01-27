package metrics

type Recorder interface {
	//Init inits the recorder
	Init()

	//RecordSpawn records the spawn of an actor
	RecordSpawn(pid string)

	//RecordDie records the death of an actor
	RecordDie(pid string)

	//RecordDrop records the amount of messages still in
	//an actors inbox when the actor went down
	RecordDrop(pid string, amount int)

	//RecordDropRemote records the amount of messages still
	//waiting to be sent to a remote machine when the remote
	//machine disconnects
	RecordDropRemote(machineId string, amount int)

	//RecordUnhandled records all messages that could not
	//be sent due to the actor or remote machine connection
	//already being down
	RecordUnhandled(target string)

	//RecordReceive records the reception of a message
	RecordReceive(pid string)

	//RecordReceiveRemote records the reception of a message
	//from a remote system
	RecordReceiveRemote(pid string)

	//RecordSendLocal records the sending of a message to a
	//local actor
	RecordSendLocal(target string)

	//RecordSendRemote records the sending of a message to a
	//remote actor
	RecordSendRemote(target string)
}
