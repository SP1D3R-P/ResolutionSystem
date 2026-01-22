

import chromadb
import dataclasses
import os 
import pathlib 

import numpy as np 

from transformers import DPRContextEncoder , DPRQuestionEncoder , DPRContextEncoderTokenizer , DPRQuestionEncoderTokenizer
import torch

VECTOR_DB_PATH : str = os.getenv('VECTOR_DB_PATH')
print(VECTOR_DB_PATH)
CHROMA_CLIENT : chromadb.ClientAPI = chromadb.PersistentClient(path=VECTOR_DB_PATH)


class db_filds : 

    # cls var
    _collections : dict[str,chromadb.Collection] = {}

    def __init__(self,name : str):
        if db_filds._collections.get(name,None) == None : 
            collection = CHROMA_CLIENT.get_or_create_collection(
                name=name,
                metadata={"hnsw:space": "cosine"} # use cosine distance 
            )
            db_filds._collections[name] = collection

        self._collection_name = name
        self._collection = db_filds._collections[name]

    @property
    def collection(self) -> chromadb.Collection : 
        return self._collection
    
    @property
    def name(self) -> str : 
        return self._collection_name
    

class Singleton(type):
    _instances = {}
    def __call__(cls, *args, **kwargs):
        if cls not in cls._instances:
            cls._instances[cls] = super(Singleton, cls).__call__(*args, **kwargs)
        return cls._instances[cls]
    


Encoder_Path = pathlib.Path(__file__).parent / "Tokenizer" 
# Note i'm storing encoder in same dir
encoder_ctx_name = "dpr-ctx_encoder-single-nq-base"
encoder_qn_name =  "dpr-question_encoder-single-nq-base"
try : 

    context_tokenizer : DPRContextEncoderTokenizer  = DPRContextEncoderTokenizer.from_pretrained(Encoder_Path / encoder_ctx_name)
    context_encoder =  DPRContextEncoder.from_pretrained(Encoder_Path / encoder_ctx_name)

    # Question Tokenizer and Encoder 
    question_tokenizer = DPRQuestionEncoderTokenizer.from_pretrained(Encoder_Path / encoder_qn_name)
    question_encoder = DPRQuestionEncoder.from_pretrained(Encoder_Path / encoder_qn_name)
except :  
    context_tokenizer : DPRContextEncoderTokenizer  = DPRContextEncoderTokenizer.from_pretrained(f"facebook/{encoder_ctx_name}")
    context_encoder =  DPRContextEncoder.from_pretrained(f'facebook/{encoder_ctx_name}')

    context_tokenizer.save_pretrained(Encoder_Path / encoder_ctx_name)
    context_encoder.save_pretrained(Encoder_Path / encoder_ctx_name)

    # Question Tokenizer and Encoder 
    question_tokenizer = DPRQuestionEncoderTokenizer.from_pretrained(f"facebook/{encoder_qn_name}")
    question_encoder = DPRQuestionEncoder.from_pretrained(f'facebook/{encoder_qn_name}')
    
    question_tokenizer.save_pretrained(Encoder_Path / encoder_qn_name)
    context_encoder.save_pretrained(Encoder_Path / encoder_qn_name)

DEVICE = 'cpu' # TODO :: Change to cuda
# 'gpu' if torch.cuda.is_available() else 'cpu'

context_encoder.to(DEVICE)  
question_encoder.to(DEVICE)

def str_ctx_embedding(ctx : str ) -> np.ndarray: 
    token = context_tokenizer(ctx , return_tensors = 'pt', padding=True, truncation=True,max_length=500).to(DEVICE)
    embeding = None 
    with torch.no_grad() : 
        output = context_encoder(**token)
        embeding = output.pooler_output.cpu().numpy() 
    return embeding
            
def str_qn_embedding(ctx : str ) -> np.ndarray: 
    token = question_tokenizer(ctx , return_tensors = 'pt').to(DEVICE)
    embeding = None 
    with torch.no_grad() : 
        output = question_encoder(**token)
        embeding = output.pooler_output.cpu().numpy() 
    return embeding
            

# TODO need to see which cuda version i'm using in global version 
# pretty sure i'm using 12.6 globally ? or might be in conda 
# or might be using 11.8 for tf ; 
# pytorch with cuda12.6 
# pip3 install torch torchvision --index-url https://download.pytorch.org/whl/cu126



# i might use sklearns tfidf [let's decide latter ]
class MiniTfidf:
    ...
            
