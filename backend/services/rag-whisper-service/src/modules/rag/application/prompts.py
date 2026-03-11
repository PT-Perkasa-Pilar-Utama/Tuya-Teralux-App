"""
Prompt Templates for RAG operations.

Centralized prompt templates following DRY principle.
All prompts used across services should be defined here.
"""
from string import Template


class RAGPrompts:
    """
    Collection of prompt templates for RAG operations.

    Using string.Template for safe, readable template substitution.
    """

    # Query/Chat Prompts
    QUERY = Template(
        "Context:\n$context\n\n"
        "Question: $question\n\n"
        "Answer the question based ONLY on the context provided. "
        "If the answer cannot be found in the context, state that you don't know."
    )

    CHAT = Template(
        "Context:\n$context\n\n"
        "User Question: $question\n\n"
        "Respond as a helpful assistant. Use the context when relevant, "
        "but you can also provide general knowledge when appropriate."
    )

    # Translation Prompts
    TRANSLATE = Template(
        "Translate the following text to $target_lang:\n\n$text"
    )

    # Summarization Prompts
    SUMMARIZE = Template(
        "Summarize the following text concisely:\n\n$text"
    )

    SUMMARIZE_DETAILED = Template(
        "Provide a detailed summary of the following text, "
        "capturing all key points:\n\n$text"
    )

    # Device Control Prompts
    CONTROL = Template(
        "Analyze this user command for device control: $command\n\n"
        "Return the intended action and device logic. "
        "Identify: 1) The action (on/off/dim/color/etc.), "
        "2) The target device (if mentioned), "
        "3) Any parameters (brightness level, color, etc.)"
    )

    # Refinement Prompts
    REFINE_ANSWER = Template(
        "Refine the following answer to be more clear and concise:\n\n$answer"
    )

    # Context Building
    CONTEXT_SEPARATOR = "\n\n---\n\n"

    @classmethod
    def build_context(cls, documents: list[str]) -> str:
        """
        Build a context string from multiple documents.

        Args:
            documents: List of document contents.

        Returns:
            Combined context string.
        """
        return cls.CONTEXT_SEPARATOR.join(documents)

    @classmethod
    def format_query(cls, context: str, question: str) -> str:
        """
        Format a query prompt with context and question.

        Args:
            context: The retrieved context.
            question: The user's question.

        Returns:
            Formatted query prompt.
        """
        return cls.QUERY.substitute(context=context, question=question)

    @classmethod
    def format_chat(cls, context: str, question: str) -> str:
        """
        Format a chat prompt with context and question.

        Args:
            context: The retrieved context.
            question: The user's question.

        Returns:
            Formatted chat prompt.
        """
        return cls.CHAT.substitute(context=context, question=question)

    @classmethod
    def format_translate(cls, text: str, target_lang: str) -> str:
        """
        Format a translation prompt.

        Args:
            text: The text to translate.
            target_lang: The target language.

        Returns:
            Formatted translation prompt.
        """
        return cls.TRANSLATE.substitute(text=text, target_lang=target_lang)

    @classmethod
    def format_summarize(cls, text: str, detailed: bool = False) -> str:
        """
        Format a summarization prompt.

        Args:
            text: The text to summarize.
            detailed: Whether to use detailed summarization.

        Returns:
            Formatted summarization prompt.
        """
        template = cls.SUMMARIZE_DETAILED if detailed else cls.SUMMARIZE
        return template.substitute(text=text)

    @classmethod
    def format_control(cls, command: str) -> str:
        """
        Format a device control analysis prompt.

        Args:
            command: The user's control command.

        Returns:
            Formatted control prompt.
        """
        return cls.CONTROL.substitute(command=command)
