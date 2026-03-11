"""
In-Memory Session Repository Implementation.

Implements SessionRepository for RAGSession persistence.
For production use, replace with database-backed implementation.
"""
from typing import List, Optional
from datetime import datetime

from src.modules.rag.domain.repositories import SessionRepository
from src.modules.rag.domain.aggregates import RAGSession


class InMemorySessionRepository(SessionRepository):
    """
    In-memory session repository implementation.

    Provides session persistence using in-memory storage.
    Suitable for development/testing. For production, use a
    database-backed implementation (Redis, PostgreSQL, etc.).
    """

    def __init__(self):
        """Initialize the in-memory session repository."""
        self._sessions: dict[str, RAGSession] = {}
        self._user_index: dict[str, list[str]] = {}  # user_id -> [session_ids]

    async def save(self, session: RAGSession) -> str:
        """
        Save a session to the repository.

        Args:
            session: The session to save.

        Returns:
            The ID of the saved session.
        """
        self._sessions[session.session_id] = session

        # Index by user_id
        if session.user_id not in self._user_index:
            self._user_index[session.user_id] = []
        if session.session_id not in self._user_index[session.user_id]:
            self._user_index[session.user_id].append(session.session_id)

        return session.session_id

    async def find_by_id(self, session_id: str) -> Optional[RAGSession]:
        """
        Find a session by its ID.

        Args:
            session_id: The session ID to search for.

        Returns:
            The session if found, None otherwise.
        """
        return self._sessions.get(session_id)

    async def find_by_user(self, user_id: str) -> List[RAGSession]:
        """
        Find all sessions for a specific user.

        Args:
            user_id: The user ID to search for.

        Returns:
            List of sessions owned by the user.
        """
        session_ids = self._user_index.get(user_id, [])
        return [
            self._sessions[sid]
            for sid in session_ids
            if sid in self._sessions
        ]

    async def find_active_sessions(
        self,
        user_id: str,
        since: Optional[datetime] = None
    ) -> List[RAGSession]:
        """
        Find active sessions for a user.

        Args:
            user_id: The user ID to search for.
            since: Optional cutoff datetime for "active" sessions.

        Returns:
            List of active sessions.
        """
        sessions = await self.find_by_user(user_id)

        if since is None:
            # Default: sessions updated in the last 24 hours
            since = datetime.utcnow().replace(hour=0, minute=0, second=0, microsecond=0)

        return [
            session for session in sessions
            if session.updated_at >= since
        ]

    async def delete(self, session_id: str) -> bool:
        """
        Delete a session from the repository.

        Args:
            session_id: The ID of the session to delete.

        Returns:
            True if deleted, False if session didn't exist.
        """
        if session_id not in self._sessions:
            return False

        session = self._sessions[session_id]

        # Remove from user index
        if session.user_id in self._user_index:
            if session_id in self._user_index[session.user_id]:
                self._user_index[session.user_id].remove(session_id)

        # Remove session
        del self._sessions[session_id]
        return True

    async def update_last_accessed(
        self,
        session_id: str,
        timestamp: Optional[datetime] = None
    ) -> bool:
        """
        Update the last accessed timestamp for a session.

        Args:
            session_id: The session ID to update.
            timestamp: Optional timestamp (defaults to now).

        Returns:
            True if updated, False if session didn't exist.
        """
        if session_id not in self._sessions:
            return False

        session = self._sessions[session_id]
        session.updated_at = timestamp or datetime.utcnow()
        return True
