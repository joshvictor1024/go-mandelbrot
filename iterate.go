package main

func iterate(cre, cim float64, maxIt uint) uint {
	var zre, zim float64 = 0, 0
	var it uint = 0
	for ; zre*zre+zim*zim < 4 && it < maxIt; it += 1 {
		// z = z ^ 2 + c
		copyZre := zre
		zre = zre*zre - zim*zim + cre
		zim = copyZre*zim*2 + cim
	}
	return it
}
