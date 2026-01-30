import grpc
from concurrent import futures
import numpy as np 
import pydoc
import requests
import pathlib

from PyRPC.proto.common_pb2_grpc import (
    add_UserFeatureSeviceServicer_to_server,
    add_DevFeatureServiceServicer_to_server,

    add_IssueIntermideateServiceServicer_to_server,
    add_UserIssueServiceServicer_to_server,
    add_DevIssueServiceServicer_to_server
)
from PyRPC.services.Features_Server import UserFeatureService , DevFeatureService
from PyRPC.services.IssueServer import IssueIntermideateService , UserIssueService , DevIssueService
from PyRPC.proto.common_pb2 import Feature
import os 

def serve():
    PORT : int = int(os.getenv('PY_PORT'))
    server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

    # if not (pathlib.Path(__file__).parent / os.getenv('VECTOR_DB_PATH')).exists():
    dev_f : DevFeatureService = DevFeatureService()
    dev_f.AddFeaturePy(
        name= 'numpy.Zeros',
        desc="Forms Array with zero"
    )
    dev_f.AddFeaturePy(
        name= 'numpy.argmax',
        desc="Max of Given Iter Arg"
    )
    dev_f.AddFeaturePy(
        name= 'numpy.argmin',
        desc="Min of Given Iter Arg"
    )
    dev_f.AddFeaturePy(
        name= 'requests.get',
        desc="Request using Get Method"
    )
    dev_f.AddFeaturePy(
        name= 'requests.post',
        desc="Request using POST Method"
    )
    # print("Data Inserted")
    print(dev_f._db.collection.get())
    
    # Feature part 
    add_UserFeatureSeviceServicer_to_server(UserFeatureService(),server=server)
    add_DevFeatureServiceServicer_to_server(DevFeatureService(),server=server)

    add_IssueIntermideateServiceServicer_to_server(IssueIntermideateService(0.4526),server=server)
    add_UserIssueServiceServicer_to_server(UserIssueService(),server=server)
    add_DevIssueServiceServicer_to_server(DevIssueService(),server=server)


    server.add_insecure_port(f"localhost:{PORT}")
    server.start()
    print(f"Service running on port {PORT}")
    server.wait_for_termination()


