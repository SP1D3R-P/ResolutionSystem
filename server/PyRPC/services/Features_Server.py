

from PyRPC.proto.common_pb2 import Feature  
from PyRPC.proto.common_pb2_grpc import UserFeatureSeviceServicer

import PyRPC.proto.common_pb2_grpc as pb_g
import PyRPC.proto.common_pb2 as pb 

import grpc


import uuid

from typing import Any , Generator
from typing_extensions import Self

from . import SharedData

import chromadb

FEATURE_VEC_DB_NAME = "Features"

# NOTE :: Many assumstions are made like 
#  1. we can't have features with same name 
#  2. Each features must have description 

class UserFeatureService(UserFeatureSeviceServicer):
    
    __metaclass__ = SharedData.Singleton
    _db : SharedData.db_filds = SharedData.db_filds(FEATURE_VEC_DB_NAME)  

    @staticmethod
    def _val_to_Feature(id : str , desc : str , metadata : dict ) -> pb.Feature : 
        return pb.Feature(
            id=pb.Id(id=id),
            name=pb.Name(name=metadata['Feature-Name']),
            description=pb.Description(description=desc)
        )

    def GetAllFeatures(self, request : pb.Empty  , context : grpc.RpcContext ) -> Generator[pb.Feature,None,None] :
        try : 
            all_data : dict = self._db.collection.get()
            for (id,meta_data,doc) in zip(all_data['ids'],all_data['metadatas'],all_data['documents']):
                yield self._val_to_Feature(id,doc,meta_data)
        except Exception as e : 
            raise grpc.RpcError(
                e.args
            ) # forwarding the error 


    def GetFeatureByName(self, request : pb.Name , context : grpc.RpcContext ) -> pb.Feature | grpc.RpcError :
 
        # chromadb.Where
        result = self._db.collection.get(where={'Feature-Name' : request.name})  
        if len(result['ids']) == 0 : 
            raise  grpc.RpcError(
                "Feature Name {} not found".format(request.name)
            )

        return self._val_to_Feature(
            result['ids'][0],
            result['documents'][0],
            result['metadatas'][0],
        )
    def GetFeatureByDsc(self, request : pb.Description , context : grpc.RpcContext ) -> Generator[pb.Feature,None,None] :

        embedding = SharedData.str_qn_embedding(request.description)
        result = self._db.collection.query(
            embedding=embedding
        )
        if len(result['ids']) == 0 : 
            return None 
        
        # TODO :: rank by tf-idf ?!
        
        for (id,doc,metadata) in  zip(result['ids'],result['documents'],result['metadatas']):
            yield self._val_to_Feature(
                id,
                doc,
                metadata
            )
    
######################################################
class DevFeatureService(pb_g.DevFeatureServiceServicer) :

    __metaclass__= SharedData.Singleton
    _db : SharedData.db_filds = SharedData.db_filds(FEATURE_VEC_DB_NAME)


    # for normal python call to an grpc service
    def AddFeaturePy(self : Self , name : str , desc : str  ) : 

        # Check if there any Privous Function
        result =self._db.collection.get(where={"Feature-Name":name})
        id = result['ids'][0] if len(result['ids']) == 1 else str(uuid.uuid4())

        # generating embedding 
        embeddings : chromadb.Embeddings = SharedData.str_ctx_embedding(desc) 

        # append to the database 
        self._db.collection.upsert(
            ids = id,
            documents = desc,
            embeddings = embeddings,
            metadatas= {"Feature-Name":name}   
        )

        return UserFeatureService._val_to_Feature(
            id=id,
            desc=desc,
            metadata={"Feature-Name":name}   
        )
    
    def AddFeature(self : Self , request : pb.Feature , context : grpc.RpcContext ) -> pb.Feature | ValueError : 
        return self.AddFeaturePy(request.name,request.description)
    
    def ListAllFeatureName(self, request : pb.Empty , context) -> Generator[pb.Feature,None,None]:
        result = self._db.collection.get()
        for id in result['ids'] : 
            yield id
        return 