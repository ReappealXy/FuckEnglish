export const booksPerPage = 5;

export function paginateBooks<T>(books: readonly T[], requestedPage: number) {
  const pages = Math.ceil(books.length / booksPerPage);
  const page = Math.max(0, Math.min(requestedPage, Math.max(0, pages - 1)));
  return { page, pages, books: books.slice(page * booksPerPage, (page + 1) * booksPerPage) };
}

export function selectedBookPage(books: readonly { name: string }[], selected: string) {
  return Math.floor(Math.max(0, books.findIndex(book => book.name === selected)) / booksPerPage);
}
