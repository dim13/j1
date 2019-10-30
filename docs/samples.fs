: fibonacci dup 2 < if drop 1 else dup 2 - recurse swap 1 - recurse + then ;
: fibnums for i fibonacci u. next ;
