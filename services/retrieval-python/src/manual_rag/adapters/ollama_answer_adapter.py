import ollama
from manual_rag.ports.answer_generator_port import AnswerGeneratorPort

PROMPT_TEMPLATE = """Eres un asistente técnico especializado en manuales de mantenimiento y lubricación.
Responde a la pregunta del usuario utilizando ÚNICAMENTE la información provista en los fragmentos del manual a continuación.
Si la respuesta no se encuentra en el texto provisto, indica claramente que la información no está disponible en el manual.

FRAGMENTOS DEL MANUAL:
{context}

PREGUNTA DEL USUARIO:
{query}

RESPUESTA DETALLADA (incluye referencias a los números de página si aplica):"""


class OllamaAnswerAdapter(AnswerGeneratorPort):
    def __init__(self, model: str, host: str = ""):
        self._model = model
        self._client = ollama.Client(host=host) if host else ollama

    def generate_answer(self, query: str, context: str) -> str:
        prompt = PROMPT_TEMPLATE.format(context=context, query=query)
        response = self._client.generate(model=self._model, prompt=prompt)
        return response["response"]
