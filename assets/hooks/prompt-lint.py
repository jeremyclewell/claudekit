#!/usr/bin/env python3
import sys, json
data = sys.stdin.read()
try:
    obj = json.loads(data)
    prompt = obj.get("prompt","")
except Exception:
    prompt = ""
# Simple checks: ask user to add acceptance criteria
if len(prompt) > 0 and "acceptance" not in prompt.lower():
    sys.stderr.write("Please include acceptance criteria in your prompt (exit=2 to block).\n")
    sys.exit(2)  # block & show to user
# Otherwise, inject brief guidance to context via stdout
sys.stdout.write("Consider: objective, constraints, acceptance criteria, and test plan.")
sys.exit(0)