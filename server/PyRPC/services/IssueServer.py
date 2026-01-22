
import PyRPC.proto.common_pb2_grpc as pb_g
import PyRPC.proto.common_pb2 as pb 

from typing import Any , Generator
from typing_extensions import Self


from PyRPC.services.SharedData import db_filds
import PyRPC.services.SharedData 
from PyRPC.services import SharedData
import chromadb

from . import Features_Server
import random
import os 
import grpc

from uuid import uuid4

ISSUE_VEC_DB_NAME = "Issue"

class IssueIntermideateService(pb_g.IssueIntermideateServiceServicer) : 
    
    _db : db_filds = db_filds(ISSUE_VEC_DB_NAME)
    _features : db_filds = db_filds(Features_Server.FEATURE_VEC_DB_NAME)

    def __init__(self : Self, thrashold : int ):
        super().__init__()
        self.thrashold = thrashold


    __metadata__ = SharedData.Singleton

    def CreateIssue(self : Self , request : pb.Issue , context : Any) -> pb.CreateIssueResponse:
        id : str = str(uuid4())

        try :
            embeddings = SharedData.str_ctx_embedding(request.description)
            self._db.collection.add(
                ids= id,
                embeddings=embeddings,
                metadatas={"title" : request.title.title , "status" : 0 , "feature_id" : request.related.id , "solution" : "" },
                documents=request.description
            )   
        except Exception as e : 
            print("Can't Insert due to ERROR:: ",e)
        return pb.Issue(
            id=pb.Id(id=id),
            title=pb.Title(title=request.title.title),
            related=request.related,
            description=pb.Description(description=request.description),
            status=1,
            solution=""
        )
    

    @staticmethod
    def query_to_issue(id : str ,title : str , feature : pb.Feature , desc : str , status : str , sol : str  ) -> pb.Issue:
        
        return pb.Issue (
            id = pb.Id(id=id) ,
            title=pb.Title(title=title),
            related=feature,
            description=pb.Description(description=desc),
            status=status,
            solution=sol

        )
        
    def _filterIssues(self : Self , distances : list[float] ) -> list[bool] : 
        return list(map(lambda x : x < self.thrashold , distances ))
    
    @staticmethod
    def FindSimilarIssuePy(db : db_filds , i_desc : str, feature_id : str , *,k = 10) -> dict : 
        embeddings = SharedData.str_qn_embedding(i_desc)
        return db.collection.query(
            query_embeddings=embeddings,
            where={"feature_id":feature_id},
            n_results=k
        )


    def FindSimilarIssue(self : Self , request : pb.FindSimilarIssueArgs , context : Any) -> Generator[pb.Issue,None,None] :
        query : chromadb.QueryResult  = IssueIntermideateService.FindSimilarIssuePy(
                self._db,
                request.issue.description.description,
                request.issue.related.id.id,
                k=10
            )
        filters = self._filterIssues()
        for (id,doc,metatdata,f) in zip(query['ids'][0],query['documents'][0],query['metadatas'][0],filters):
            # "title" : request.title , "status" : 0 , "feature_id" : request.related.id , "solution" : ""
            if f :
                feature : chromadb.GetResult = IssueIntermideateService._features.collection.get(ids=metatdata['feature_id'])
                feature : pb.Feature = Features_Server.UserFeatureService._val_to_Feature(
                    id = feature["ids"][0],
                    desc= feature["documents"][0],
                    metadata=feature['metadatas'][0]
                )
                yield IssueIntermideateService.query_to_issue(
                        id=id,
                        title="Some title",
                        feature=feature,
                        desc=doc,
                        status=metatdata['status'],
                        sol=metatdata['solution']
                    )
                

        



class UserIssueService(pb_g.UserIssueServiceServicer) : 
    """
        planning to shift this part to go
        that's why IssueIntermideateService service is naked?
    """
    similarity_thrashold : float = .40
    _db = db_filds(ISSUE_VEC_DB_NAME)
    __metadata__ = SharedData.Singleton

    def __init__(self):
        super().__init__()
        self.channel = grpc.insecure_channel(f'localhost:{os.getenv("PY_PORT")}')
        self.client = pb_g.IssueIntermideateServiceStub(self.channel)
        

    def PostIssue(self : Self , request : pb.Issue , context : Any ) -> pb.PostIssueResponse:
        
        result = []
        try : 
            query : chromadb.QueryResult  = IssueIntermideateService.FindSimilarIssuePy(
                self._db,
                request.description.description,
                request.related.id.id,
                k=10
            )
            filters = list(map(lambda x : x > self.similarity_thrashold , query['distances'][0] ))
            for (id,doc,metatdata,f) in zip(query['ids'][0],query['documents'][0],query['metadatas'][0],filters):
                # "title" : request.title , "status" : 0 , "feature_id" : request.related.id , "solution" : ""
                if f :
                    feature : chromadb.GetResult = IssueIntermideateService._features.collection.get(ids=metatdata['feature_id'])
                    feature : pb.Feature = Features_Server.UserFeatureService._val_to_Feature(
                        id = feature["ids"][0],
                        desc= feature["documents"][0],
                        metadata=feature['metadatas'][0]
                    )
                    result.append(IssueIntermideateService.query_to_issue(
                            id=id,
                            title="Some title",
                            feature=feature,
                            desc=doc,
                            status=metatdata['status'],
                            sol=metatdata['solution']
                        )
                    )
        
        except Exception as e  : 
            print("ERROR ::",e)

        # try :
        #     query = self.client.FindSimilarIssue(
        #         pb.FindSimilarIssueArgs(
        #             issue=request,
        #             k=pb.Limit(k=10)
        #         ),
        #         timeout=1.0
        #     ) 
        #     for val in query :
        #         result.append(val)
        # except Exception as err:
        #     print(err)
        # print("I'm Here in post issue",result)

        if len(result) < 1 : 
            # creating new issue 
            id : str = str(uuid4())
            try :
                embeddings = SharedData.str_ctx_embedding(request.description.description)
                self._db.collection.add(
                    ids= id,
                    embeddings=embeddings,
                    metadatas={"title" : request.title.title , "status" : 0 , "feature_id" : request.related.id.id , "solution" : "" },
                    documents=request.description.description
                )   
                created = pb.Issue(
                    id=pb.Id(id=id),
                    title=pb.Title(title=request.title.title),
                    related=request.related,
                    description=request.description,
                    status=0,
                    solution=""
                )
            except Exception as e : 
                print("Can't Insert due to ERROR:: ",e)
            return pb.PostIssueResponse(
                created=pb.CreateIssueResponse(issue=created),
                isCreated=True
            )
            # return None

        # this one will be done based on tf-idf based 
        def filter_best() :
            ...

        # result = filter_best(result) 
        
        return pb.PostIssueResponse(
            issues= pb.IssueList(issues=result),
            isCreated=False
        )
    
    def GetIssuesByFeatureName(self : Self, request : pb.Name, context : Any ) -> Generator[pb.Issue,None,None]:
        return super().GetIssuesByFeatureName(request, context)
    
    def GetIssueByTitle(self : Self, request : pb.Title, context : Any ) -> pb.Issue:
        return super().GetIssueByTitle(request, context)
    
    def GetIssueById(self : Self, request : pb.Id , context) -> pb.Issue:
        return super().GetIssueById(request, context)
    

class DevIssueService(pb_g.DevIssueServiceServicer) :
    _db : db_filds = db_filds(ISSUE_VEC_DB_NAME)

    def GetAllIssues(self : Self, request : pb.Empty , context : Any ) -> Generator[pb.Issue,None,None]:
        ...
    
    def GetIssuesByFeatureName(self : Self, request : pb.Name, context : Any ) -> Generator[pb.Issue,None,None]:
        return super().GetIssuesByFeatureName(request, context)
    
    def GetPendingIssues(self : Self, request : pb.Empty , context : Any) -> Generator[pb.Issue,None,None]:
        return super().GetPendingIssues(request, context)
    
