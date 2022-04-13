build:
	garble -literals -tiny -seed=random build -ldflags="-s -w -H=windowsgui" -o simple.exe .