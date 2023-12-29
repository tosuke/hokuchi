VERSION 0.7

ipxe-builder:
    FROM ubuntu:22.04
    RUN \
        apt update && \
        apt install -y --no-install-recommends \
            ca-certificates \
            git \
            make \
            perl \
            gcc-x86-64-linux-gnu \
            gcc-aarch64-linux-gnu \
            libc6-dev \
            mtools
    RUN git clone --filter=tree:0 https://github.com/ipxe/ipxe.git
    WORKDIR /ipxe
    COPY ./embed.ipxe ./src/embed.ipxe

ipxe-amd64:
    FROM +ipxe-builder
    RUN make -j -C src CROSS=x86_64-linux-gnu- EMBED=embed.ipxe bin-x86_64-efi/ipxe.efi
    SAVE ARTIFACT /ipxe/src/bin-x86_64-efi/ipxe.efi AS LOCAL ./assets/boot_amd64.efi

ipxe-arm64:
    FROM +ipxe-builder
    RUN make -j -C src CROSS=aarch64-linux-gnu- EMBED=embed.ipxe bin-arm64-efi/ipxe.efi
    SAVE ARTIFACT /ipxe/src/bin-arm64-efi/ipxe.efi AS LOCAL ./assets/boot_arm64.efi
