# go-ase

To be able to compile this you have to have the C-header files and the libraries.
Therefore please install the SDK according to the [Open Server Unix Install Guide].
Afterwards you have to setup the following environment variables.

```
export CGO_CFLAGS="-I/opt/sap/OCS-16_0/include"
export CGO_LDFLAGS="-L/opt/sap/OCS-16_0/lib -lsybct_r64 -lsybcs_r64"
export LD_LIBRARY_PATH="/opt/sap/OCS-16_0/lib"
```

If you installed the SDK to a different location you have to adjust the variables.
Finally compile the driver and the examples by running `make all`.

[Open Server Unix Install Guide]: https://help.sap.com/viewer/882ef48c7e9c4d6e845d98f34378db40/16.0.3.2/en-US
