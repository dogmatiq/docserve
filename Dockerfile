FROM scratch
COPY artifacts/build/release/linux/amd64/browser /bin/browser
ENTRYPOINT ["/bin/browser"]
