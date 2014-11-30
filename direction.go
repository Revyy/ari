package ari

import "fmt"

func (c *Client) CheckDirection(dir *string) error {
	if *dir != "in" && *dir != "out" && *dir != "both" {
		if *dir == "" {
			*dir = "in"
			return nil
		}
		return fmt.Errorf("Not a viable option for direction.")
	}
	return nil
}
