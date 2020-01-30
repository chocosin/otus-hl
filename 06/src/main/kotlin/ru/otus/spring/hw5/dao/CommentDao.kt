package ru.otus.spring.hw5.dao

interface CommentDao {
    fun addComment(bookId: Long, text: String): Long
    fun removeComment(commentId: Long)
}
