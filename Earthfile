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
            gcc \
            gcc-aarch64-linux-gnu \
            libc6-dev \
            mtools \
            syslinux \
            isolinux
    RUN git clone --filter=tree:0 https://github.com/ipxe/ipxe.git
    WORKDIR /ipxe
    COPY ./embed.ipxe ./src/embed.ipxe

ipxe-amd64:
    FROM +ipxe-builder
    RUN make -j -C src EMBED=embed.ipxe bin-x86_64-efi/ipxe.efi
    SAVE ARTIFACT /ipxe/src/bin-x86_64-efi/ipxe.efi AS LOCAL ./boot_x86_64.efi

ipxe-aarch64:
    FROM +ipxe-builder
    RUN make -j -C src CROSS=aarch64-linux-gnu- EMBED=embed.ipxe bin-arm64-efi/ipxe.efi
    SAVE ARTIFACT /ipxe/src/bin-arm64-efi/ipxe.efi AS LOCAL ./boot_aarch64.efi
