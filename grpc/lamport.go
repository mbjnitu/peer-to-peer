package ping

func IncrementLamport(lamport int32) int32 {
	return lamport + 1
}

func SyncLamport(curLamport int32, newLamport int32) int32 {
	if curLamport < newLamport {
		curLamport = newLamport
	}
	return curLamport
}
