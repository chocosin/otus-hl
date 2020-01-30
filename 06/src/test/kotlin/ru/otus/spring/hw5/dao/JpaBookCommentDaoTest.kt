package ru.otus.spring.hw5.dao

import org.assertj.core.api.Assertions.assertThat
import org.junit.jupiter.api.Test
import org.springframework.beans.factory.annotation.Autowired
import org.springframework.boot.test.autoconfigure.orm.jpa.DataJpaTest
import org.springframework.boot.test.autoconfigure.orm.jpa.TestEntityManager
import org.springframework.context.annotation.Import
import ru.otus.spring.hw5.domain.Book
import ru.otus.spring.hw5.domain.BookComment

@DataJpaTest
@Import(JpaBookCommentDao::class, JpaBookDao::class)
internal class JpaBookCommentDaoTest {
    @Autowired
    lateinit var dao: JpaBookCommentDao
    @Autowired
    lateinit var em: TestEntityManager

    @Test
    fun `adds new book comment`() {
        var book = randomBook()
        book.authors.forEach { em.persist(it) }
        book.genres.forEach { em.persist(it) }
        em.persist(book)

        val text = "new comment"
        val id = dao.addComment(book.id, text)

        book = em.find(Book::class.java, book.id)

        assertThat(book.comments).isEqualTo(listOf(BookComment(id = id, comment = text, book = book)))
    }
}
