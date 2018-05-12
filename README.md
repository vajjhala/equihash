### Equihash

Most existing implementations of proof of work algorithms rely heavily on computational hardness, but because of the large disparity in CPU architectures this has made them prey to GPU-, ASIC and botnet-equipped user. Making it very hard for modest desktops to mine cryptocurrencies. Therefore there is a need for new algorithms that can address this issue. The solution is to rely on memory intensive computations.
Equihash is an asymmetric proof of work based on a computationally hard problem which requires a great deal of memory. It was proposed by Alex Biryukov and Dmitry Khovratovich. Original paper can be found [here](http://www.ledgerjournal.org/ojs/index.php/ledger/article/view/48)

#### Implementation

Equihash is based on the generalised birthday problem whose solution is given by Wagner's algorithm

The user is asked to provide the the parameters n, k and d. Where n is the length of the bit string. 2^k are the number of hashes that are XORed to zero. And d is difficulty filter.

Install the Go package and simply run the binary as: `equihash ‐n 48 ‐k 5 ‐d 5` to see the program in action.
