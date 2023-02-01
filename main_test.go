package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestArchive(t *testing.T) {
	arc, err := Archive("internal/testdata/basic/migrations")
	require.NoError(t, err)
	exp := `-- 20230201094614.sql --
create table users (
  id int primary key,
  name varchar(255),
  about text
);
-- atlas.sum --
h1:X97sBPjOeiRWeoEdqpIpHAdzlshqOqllEKrJS9JruPo=
20230201094614.sql h1:GkisBcewe1zpcP5IGGkbbsYSKJgV+fzlDVal+FKEjoE=
`
	require.Equal(t, exp, arc)
}
