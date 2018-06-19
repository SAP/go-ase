all:
	rm -f bridge.o libbridge.a &&\
	gcc -c bridge.c -Iincludes &&\
	ar cru libbridge.a bridge.o &&\
	go build -a &&\
	rm -f bridge.o libbridge.a
