.PHONY: all clean

X86_CFLAGS := -arch x86_64
X86_CFLAGS += -mavx
X86_CFLAGS += -mavx2
X86_CFLAGS += -mno-bmi

ARM_CFLAGS := -march=armv8-a+fp+simd
ARM_CFLAGS += -Itools/sse2neon

CFLAGS := -mno-red-zone
CFLAGS += -fno-asynchronous-unwind-tables
CFLAGS += -fno-stack-protector
CFLAGS += -fno-exceptions
CFLAGS += -fno-builtin
CFLAGS += -fno-rtti
CFLAGS += -nostdlib
CFLAGS += -O3

NATIVE_ASM := $(wildcard native/*.S)
NATIVE_SRC := $(wildcard native/*.h)
NATIVE_SRC += $(wildcard native/*.c)

all: native_amd64.s native_arm64.s

clean:
	rm -vf native_amd64.s native_arm64.s output/*.s

native_amd64.s: ${NATIVE_SRC} native_amd64.go
	mkdir -p output
	clang ${X86_CFLAGS} ${CFLAGS} -S -o output/native_amd64.s native/native.c
	python3 tools/asm2asm/asm2asm.py native_amd64.s output/native_amd64.s
	asmfmt -w native_amd64.s

native_arm64.s: ${NATIVE_SRC} native_arm64.go
	mkdir -p output
	clang ${ARM_CFLAGS} ${CFLAGS} -S -o output/native_arm64.s native/native.c
	nocgo native_arm64.s output/native_arm64.s
	asmfmt -w native_arm64.s
