from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
from sentence_transformers import SentenceTransformer
from qdrant_client import QdrantClient

app = FastAPI(title="RAG Engine Internal API")

client = QdrantClient(host="localhost", port=6333)
embedding_model = SentenceTransformer("BAAI/bge-m3")

class SearchQuery(BaseModel):
    query: str
    top_k: int = 3

@app.post("/search")
def search_context(data: SearchQuery):
    try:
        vector = embedding_model.encode(data.query).tolist()
        response = client.query_points(
            collection_name="manuales_tecnicos",
            query=vector,
            limit=data.top_k
        )

        results = []
        for point in response.points:
            results.append({
                "page": point.payload.get("page"),
                "source": point.payload.get("source"),
                "text": point.payload.get("text")
            })
        return {"results": results}
    except Exception as e:
        raise HTTPException(status_code=500, detail=str(e))

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)