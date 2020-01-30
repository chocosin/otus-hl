package ru.otus.spring.hw5.domain

import javax.persistence.*

@Entity
@Table(name = "book_comments")
class BookComment(
        @Id
        @Column(name = "comment_id")
        @GeneratedValue(strategy = GenerationType.IDENTITY)
        var id: Long,

        var comment: String,

        @ManyToOne(
                optional = false,
                fetch = FetchType.LAZY
        )
        @JoinColumn(name = "book_id")
        var book: Book
)
