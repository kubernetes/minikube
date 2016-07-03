package xhyve

/*
#include <stdio.h>
#include <uuid/uuid.h>

char uuid_str[37];

extern inline char* uuidgen() {
	// typedef unsigned char uuid_t;
	uuid_t uuid;

	// generate
	uuid_generate_random(uuid);

	// unparse (to string)
	uuid_unparse_upper(uuid, uuid_str);

	// return printf("%s\n", uuid_str);
	return uuid_str;
}
*/
import "C"

func uuidgen() string {
	return C.GoString(C.uuidgen())
}
