package server

type Neighborhood struct {
	ID   int64
	Slug string
	Name string
}

func (s *Server) NeighborhoodBySlug(slug string) (*Neighborhood, error) {
	var n Neighborhood
	err := s.ConnPool.QueryRow(`
		select id, slug, name from neighborhoods
		where lower(slug) = lower($1)
	`, slug).Scan(&n.ID, &n.Slug, &n.Name)
	switch {
	// case err == pgx.ErrNoRows:
	case err != nil:
		return nil, err
	}

	return &n, nil
}

func (s *Server) NeighborhoodByID(id int64) (*Neighborhood, error) {
	var n Neighborhood
	err := s.ConnPool.QueryRow(`
		select id, slug, name from neighborhoods
		where id = $1
	`, id).Scan(&n.ID, &n.Slug, &n.Name)
	switch {
	// case err == pgx.ErrNoRows:
	case err != nil:
		return nil, err
	}

	return &n, nil
}
