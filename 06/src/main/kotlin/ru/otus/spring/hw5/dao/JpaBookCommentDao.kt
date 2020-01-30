package ru.otus.spring.hw5.dao

import org.springframework.stereotype.Repository
import org.springframework.transaction.annotation.Transactional
import ru.otus.spring.hw5.domain.BookComment
import javax.persistence.EntityManager
import javax.persistence.PersistenceContext

@Repository
@Transactional
class JpaBookCommentDao(
        @PersistenceContext
        private val em: EntityManager,
        private val bookDao: BookDao
) : CommentDao {
    override fun addComment(bookId: Long, text: String): Long {
        val book = bookDao.getById(bookId)
                ?: throw IllegalArgumentException("book not found")
        val bookComment = BookComment(id = 0, comment = text, book = book)
        book.comments += bookComment
        em.merge(book)
        return bookComment.id
    }

    override fun removeComment(commentId: Long) {
        em.find(BookComment::class.java, commentId)?.also {
            em.remove(it)
        }
    }
}
