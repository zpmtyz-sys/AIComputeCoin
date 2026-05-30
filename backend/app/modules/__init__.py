"""Business modules. Each module owns its tables and exposes a service.py + router.py.

Cross-module calls go ONLY through service layers (CLAUDE.md). To avoid import cycles,
cross-module service imports are done lazily inside functions.
"""
