FROM scratch

COPY tk /usr/bin/tk

ENTRYPOINT [ "/usr/bin/tk" ]

CMD ["serve"]

