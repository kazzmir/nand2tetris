load Eq16Zero.hdl,
output-file Eq16Zero.out,
compare-to Eq16Zero.cmp,
output-list in%B1.16.1 zero%B3.1.3;

set in %B0000000000000000,
eval,
output;

set in %B0000000000000001,
eval,
output;

set in %B1000000000000001,
eval,
output;

set in %B1000000000000000,
eval,
output;
