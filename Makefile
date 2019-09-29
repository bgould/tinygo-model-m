build:
	tinygo build -target=feather-m0 -o keyboard.uf2 main.go

flash:
	stty -F /dev/ttyACM0 1200 hupcl; tinygo build -target=feather-m0 -size=full -o /media/$(USER)/FEATHERBOOT/keyboard.uf2 main.go
