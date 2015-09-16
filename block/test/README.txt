We have to put tests here in their own folder so we can use
the wallet to create transactions and test the blocks.  We
COULD build transactions by hand, but this 1) tests more code,
and 2) building transactions by hand and signing them is 
a royal pain.

If we try and put these tests with the block building code we
get import loops that cannot be resolved.
