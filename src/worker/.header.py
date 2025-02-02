#======== SCAFFOLD WORKER HEADER FILE ========#
import os
import json

CONTEXT = {}

def init_context() -> None:
    global CONTEXT
    context_string = os.getenv('SCAFFOLD_CONTEXT')
    if context_string:
        CONTEXT = json.loads(context_string)

def set_context(key: str, val: str) -> None:
    global CONTEXT
    CONTEXT[key] = val
    with open('.context.json', 'w', encoding='utf-8') as context_file:
        json.dump(CONTEXT, context_file)

def get_context(key: str) -> str:
    global CONTEXT
    if key in CONTEXT:
        return CONTEXT[key]
    return ""

init_context()
