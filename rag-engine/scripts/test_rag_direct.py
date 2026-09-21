import time
import requests

start_time = time.time()

url = "http://localhost:8000/search"
payload = {
    "question": "hola",
    "top_k": 3
}

print("Enviando consulta al motor RAG (Python + Qdrant)...")
try:
    # Se remueve timeout para esperar la respuesta sin importar cuánto tarde
    response = requests.post(url, json=payload)
    elapsed = time.time() - start_time
    
    print(f"\n✅ Respuesta recibida en {elapsed:.2f} segundos!")
    print("\nFragmentos encontrados:")
    for idx, doc in enumerate(response.json().get("results", []), 1):
        print(f"\n--- Resultado {idx} ---")
        print(doc.get("text")[:200] + "...")
except Exception as e:
    print(f"\n❌ Error al conectar con el motor RAG: {e}")