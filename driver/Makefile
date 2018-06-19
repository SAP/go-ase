all:
	export LD_LIBRARY_PATH=/opt/sap/OCS-16_0/lib &&\
	export CGO_LDFLAGS="-L/opt/sap/OCS-16_0/lib -lsybct_r64 -lsybcs_r64" &&\
	rm -f bridge.o libbridge.a &&\
	gcc -c bridge.c -Iincludes &&\
	ar cru libbridge.a bridge.o &&\
	go build -a &&\
	rm -f bridge.o libbridge.a
