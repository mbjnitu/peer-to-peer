package ping

func IncrementLamport(lamport int) int {
	return lamport + 1
}

func SyncLamport(curLamport int, newLamport int) int {
	if curLamport < newLamport {
		curLamport = newLamport
	}
	return curLamport
}
