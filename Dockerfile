FROM scratch

ADD kubevirt-custom-hook /custom-hook

ENTRYPOINT ["/custom-hook"]
