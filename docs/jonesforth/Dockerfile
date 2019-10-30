FROM debian
RUN apt-get update && apt-get install -y libc6-dev-i386
COPY jonesforth.* /
RUN gcc -m32 -nostdlib -static -Wl,-Ttext,0 -Wl,--build-id=none -o jonesforth jonesforth.S
CMD cat jonesforth.fs - | ./jonesforth
# NOTE requires SYS_RAWIO
