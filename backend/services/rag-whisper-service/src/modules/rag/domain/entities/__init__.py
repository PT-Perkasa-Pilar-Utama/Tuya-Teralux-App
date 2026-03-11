"""
Entities module - re-exports canonical DocumentDomain from ports.

For DDD consistency, DocumentDomain is defined in ports/ as the
canonical domain model. This module provides backward compatibility.
"""
from src.modules.rag.domain.ports.document import DocumentDomain

__all__ = ["DocumentDomain"]
