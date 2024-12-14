FROM --platform=linux/amd64 golang:1.22.8
COPY ../cmd/accrual/accrual_linux_amd64 /accrual
CMD /accrual -a "${ACCRUAL_ADDRESS}"