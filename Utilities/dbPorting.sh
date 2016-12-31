echo "Porting DB"
cd DatabasePorter
go build
./DatabasePorter level
./DatabasePorter level
./DatabasePorter level
cd ..
cd DatabaseIntegrityCheck
./DatabaseIntegrityCheck level ../DatabasePorter/database/ldb/FactoidLevel-Import.db
cd ..
echo "Done porting DB"